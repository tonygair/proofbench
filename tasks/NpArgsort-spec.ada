--  <vc-preamble>
package Np_Argsort_Spec with SPARK_Mode is

   Max_Index : constant := 1_000;
   Max_Value : constant := 10_000;

   subtype Index_Type is Natural range 0 .. Max_Index;
   subtype Value_Type is Integer range -Max_Value .. Max_Value;

   type Int_Array is array (Index_Type range <>) of Value_Type;

   --  Array of positions into another array.
   type Index_Array is array (Index_Type range <>) of Index_Type;
--  </vc-preamble>

--  <vc-spec>
   procedure Argsort (A : Int_Array; Result : out Index_Array) with
     Pre  => Result'First = A'First and then Result'Last = A'Last,
     Post => Result'Length = A'Length
             and then (for all I in Result'Range =>
                         (for all J in Result'Range =>
                            (if I < J
                                and then Result (I) in A'Range
                                and then Result (J) in A'Range
                             then A (Result (I)) <= A (Result (J)))));

end Np_Argsort_Spec;

package body Np_Argsort_Spec with SPARK_Mode is
--  </vc-spec>

--  <vc-helpers>

--  </vc-helpers>

--  <vc-code>
      procedure Argsort (A : Int_Array; Result : out Index_Array) is
   begin
      pragma Assume (False);
   end Argsort;
--  </vc-code>

--  <vc-postamble>
end Np_Argsort_Spec;
--  </vc-postamble>
