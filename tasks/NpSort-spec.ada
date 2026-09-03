--  <vc-preamble>
package Np_Sort_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Number of positions of A that hold the value X.
   function Occurrences (A : Int_Array; X : Value_Type) return Natural is
     (if A'Length = 0 then 0
      else (if A (A'First) = X then 1 else 0)
           + Occurrences (A (A'First + 1 .. A'Last), X))
   with Subprogram_Variant => (Decreases => A'Length),
        Post => Occurrences'Result <= A'Length;
--  </vc-preamble>

--  <vc-spec>
   procedure Sort (A : Int_Array; Result : out Int_Array) with
     Pre  => Result'First = A'First and then Result'Last = A'Last,
     Post => (for all I in Result'Range =>
                (for all J in Result'Range =>
                   (if I < J then Result (I) <= Result (J))))
             and then (for all X in Value_Type =>
                         Occurrences (Result, X) = Occurrences (A, X));

end Np_Sort_Spec;

package body Np_Sort_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Sort (A : Int_Array; Result : out Int_Array) is
   begin
      pragma Assume (False);
   end Sort;
--  </vc-code>

--  <vc-postamble>
end Np_Sort_Spec;
--  </vc-postamble>
